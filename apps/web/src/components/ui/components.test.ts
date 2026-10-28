import { describe, expect, it } from "vitest";
import { mount } from "@vue/test-utils";
import BaseButton from "./BaseButton.vue";
import BaseField from "./BaseField.vue";
import BasePanel from "./BasePanel.vue";
import BaseSkeleton from "./BaseSkeleton.vue";
import EmptyState from "./EmptyState.vue";

describe("BaseButton", () => {
  it("expone estado de carga con aria-busy y deshabilitado", async () => {
    const wrapper = mount(BaseButton, { props: { loading: false } });
    expect(wrapper.attributes("aria-busy")).toBe("false");
    expect(wrapper.attributes("disabled")).toBeUndefined();
    await wrapper.setProps({ loading: true });
    expect(wrapper.attributes("aria-busy")).toBe("true");
    expect(wrapper.attributes("disabled")).toBeDefined();
  });

  it("usa la variante ghost cuando se pide", () => {
    const wrapper = mount(BaseButton, { props: { variant: "ghost" } });
    expect(wrapper.classes().join(" ")).toContain("hover:bg-neutral-800");
  });
});

describe("BaseField", () => {
  it("asocia label, input y mensaje de error", () => {
    const wrapper = mount(BaseField, {
      props: { id: "correo", label: "Correo", modelValue: "", error: "campo obligatorio" },
    });
    const label = wrapper.get("label");
    expect(label.attributes("for")).toBe("correo");
    expect(wrapper.get("input").attributes("aria-invalid")).toBe("true");
    expect(wrapper.get("input").attributes("aria-describedby")).toBe("correo-error");
    expect(wrapper.get("p").attributes("id")).toBe("correo-error");
  });

  it("sin error no marca aria-invalid ni aria-describedby", () => {
    const wrapper = mount(BaseField, {
      props: { id: "correo", label: "Correo", modelValue: "" },
    });
    expect(wrapper.get("input").attributes("aria-invalid")).toBe("false");
    expect(wrapper.get("input").attributes("aria-describedby")).toBeUndefined();
  });

  it("emite update:modelValue al escribir", async () => {
    const wrapper = mount(BaseField, {
      props: { id: "correo", label: "Correo", modelValue: "" },
    });
    await wrapper.get("input").setValue("hola@ejemplo.com");
    expect(wrapper.emitted("update:modelValue")?.[0]).toEqual(["hola@ejemplo.com"]);
  });
});

describe("BasePanel", () => {
  it("vincula el titulo al encabezado accesible", () => {
    const wrapper = mount(BasePanel, { props: { title: "Bandeja" }, slots: { default: "<p>cuerpo</p>" } });
    const section = wrapper.get("section");
    const labelledBy = section.attributes("aria-labelledby");
    expect(labelledBy).toBeDefined();
    expect(wrapper.get("h2").attributes("id")).toBe(labelledBy);
  });
});

describe("BaseSkeleton", () => {
  it("anuncia estado de carga solo a lectores de pantalla", () => {
    const wrapper = mount(BaseSkeleton);
    expect(wrapper.get('[role="status"]').text()).toContain("Cargando");
    expect(wrapper.findAll('[aria-hidden="true"]').length).toBeGreaterThan(0);
  });
});

describe("EmptyState", () => {
  it("muestra titulo y descripcion", () => {
    const wrapper = mount(EmptyState, {
      props: { title: "Sin mensajes", description: "Cuando recibas un sobre aparecera aqui." },
    });
    expect(wrapper.text()).toContain("Sin mensajes");
    expect(wrapper.text()).toContain("aparecera aqui");
  });
});
